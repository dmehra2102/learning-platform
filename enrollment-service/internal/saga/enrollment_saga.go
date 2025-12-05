package saga

import (
	"context"
	"fmt"
	"time"

	"github.com/dmehra2102/learning-platform/enrollment-service/internal/domain"
	"github.com/dmehra2102/learning-platform/enrollment-service/internal/repository"
	"github.com/dmehra2102/learning-platform/shared/pkg/kafka"
	pb_course "github.com/dmehra2102/learning-platform/shared/proto/course"
	pb_payment "github.com/dmehra2102/learning-platform/shared/proto/payment"
	"github.com/google/uuid"
	"go.uber.org/zap"
	grpcLib "google.golang.org/grpc"
)

type EnrollmentRequest struct {
	UserID       string
	CourseID     string
	Amount       float64
	PaymentToken string
}

type EnrollmentSagaOrchestrator struct {
	enrollmentRepo repository.EnrollmentRepository
	paymentConn    *grpcLib.ClientConn
	courseConn     *grpcLib.ClientConn
	kafkaProducer  *kafka.Producer
	logger         *zap.Logger
}

func NewEnrollmentSagaOrchestrator(
	enrollmentRepo repository.EnrollmentRepository,
	paymentConn *grpcLib.ClientConn,
	courseConn *grpcLib.ClientConn,
	kafkaProducer *kafka.Producer,
	logger *zap.Logger,
) *EnrollmentSagaOrchestrator {
	return &EnrollmentSagaOrchestrator{
		enrollmentRepo: enrollmentRepo,
		paymentConn:    paymentConn,
		courseConn:     courseConn,
		kafkaProducer:  kafkaProducer,
		logger:         logger,
	}
}

func (o *EnrollmentSagaOrchestrator) Execute(ctx context.Context, req EnrollmentRequest) (*domain.Enrollment, error) {
	// Step-1 : Create enrollment_id in PENDING status
	enrollment := &domain.Enrollment{
		ID:         uuid.New().String(),
		UserID:     req.UserID,
		CourseID:   req.CourseID,
		Status:     domain.StatusPending,
		AmountPaid: req.Amount,
		EnrolledAt: time.Now(),
	}

	if err := enrollment.Validate(); err != nil {
		return nil, err
	}

	if err := o.enrollmentRepo.Create(ctx, enrollment); err != nil {
		o.logger.Error("failed to create enrollment", zap.Error(err))
		return nil, err
	}

	o.logger.Info("enrollment created in PENDING status", zap.String("enrollment_id", enrollment.ID))

	// Step-2: Process payment
	paymentID, err := o.processPayment(ctx, req.UserID, req.Amount, req.PaymentToken, req.CourseID)
	if err != nil {
		o.logger.Error("payment processing failed", zap.Error(err), zap.String("enrollment_id", enrollment.ID))
		enrollment.Status = domain.StatusCancelled
		_ = o.enrollmentRepo.Update(ctx, enrollment)
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	o.logger.Info("payment processed successfully", zap.String("payment_id", paymentID))

	// Step-3: Update enrollment with payment info
	enrollment.PaymentID = paymentID
	enrollment.Status = domain.StatusActive
	if err := o.enrollmentRepo.Update(ctx, enrollment); err != nil {
		o.logger.Error("failed to update enrollment after payment", zap.Error(err))
		// Try to refund
		_ = o.refundPayment(ctx, paymentID)
		enrollment.Status = domain.StatusRefunded
		_ = o.enrollmentRepo.Update(ctx, enrollment)
		return nil, fmt.Errorf("failed to activate enrollment: %w", err)
	}

	o.logger.Info("enrollment activated", zap.String("enrollment_id", enrollment.ID))

	// Step 4: Publish enrollment event
	event := domain.EnrollmentEvent{
		EnrollmentID: enrollment.ID,
		UserID:       enrollment.UserID,
		CourseID:     enrollment.CourseID,
		Status:       enrollment.Status,
		Amount:       enrollment.AmountPaid,
		Timestamp:    time.Now(),
	}

	if err := o.kafkaProducer.PublishMessage(ctx, enrollment.ID, event); err != nil {
		o.logger.Warn("failed to publish enrollment event", zap.Error(err))
	}

	o.logger.Info("enrollment saga completed successfully", zap.String("enrollment_id", enrollment.ID))
	return enrollment, nil
}

func (o *EnrollmentSagaOrchestrator) CancelEnrollment(ctx context.Context, enrollmentID string) error {
	enrollment, err := o.enrollmentRepo.GetByID(ctx, enrollmentID)
	if err != nil {
		return err
	}

	if !enrollment.CanBeCancelled() {
		return fmt.Errorf("enrollment cannot be cancelled in status: %s", enrollment.Status)
	}

	if enrollment.PaymentID != "" {
		if err := o.refundPayment(ctx, enrollment.PaymentID); err != nil {
			o.logger.Error("failed to refund payment", zap.Error(err), zap.String("payment_id", enrollment.PaymentID))
		}
		return err
	}

	// Update enrollment status
	enrollment.Status = domain.StatusCancelled
	if err := o.enrollmentRepo.Update(ctx, enrollment); err != nil {
		return err
	}

	o.logger.Info("enrollment cancelled", zap.String("enrollment_id", enrollmentID))
	return nil
}

func (o *EnrollmentSagaOrchestrator) processPayment(ctx context.Context, userID string, amount float64, token string, courseID string) (string, error) {
	client := pb_payment.NewPaymentServiceClient(o.paymentConn)

	req := &pb_payment.ProcessPaymentRequest{
		UserId:       userID,
		Amount:       amount,
		PaymentToken: token,
		CourseId:     courseID,
	}

	resp, err := client.ProcessPayment(ctx, req)
	if err != nil {
		return "", fmt.Errorf("payment service error: %w", err)
	}

	if resp.Payment.Status != pb_payment.PaymentStatus_COMPLETED {
		return "", fmt.Errorf("payment processing failed with status: %s", resp.Payment.Status)
	}

	return resp.Payment.Id, nil
}

func (o *EnrollmentSagaOrchestrator) refundPayment(ctx context.Context, paymentID string) error {
	client := pb_payment.NewPaymentServiceClient(o.paymentConn)

	req := &pb_payment.RefundPaymentRequest{
		Id:     paymentID,
		Reason: "Enrollment cancellation",
	}

	resp, err := client.RefundPayment(ctx, req)
	if err != nil {
		return fmt.Errorf("payment service error: %w", err)
	}

	if resp.Status != pb_payment.PaymentStatus_COMPLETED {
		return fmt.Errorf("refund processing failed with status: %w", err)
	}

	return nil
}

func (o *EnrollmentSagaOrchestrator) ValidateCourse(ctx context.Context, courseID string) (bool, error) {
	client := pb_course.NewCourseServiceClient(o.courseConn)

	req := &pb_course.GetCourseRequest{
		Id: courseID,
	}

	resp, err := client.GetCourse(ctx, req)
	if err != nil {
		return false, fmt.Errorf("course service error: %w", err)
	}

	return resp != nil && resp.Course != nil, nil
}
