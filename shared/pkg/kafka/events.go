package kafka

import "time"

const (
	TopicUserRegistered    = "user.registered"
	TopicCourseCreated     = "course.created"
	TopicCoursePublished   = "course.published"
	TopicEnrollmentStarted = "enrollment.started"
	TopicPaymentProcessed  = "payment.processed"
	TopicPaymentFailed     = "payment.failed"
	TopicEnrollmentSuccess = "enrollment.success"
	TopicEnrollmentFailed  = "enrollment.failed"
	TopicProgressUpdated   = "progress.updated"
	TopicLessonCompleted   = "lesson.completed"
	TopicCourseCompleted   = "course.completed"
	TopicReviewCreated     = "review.created"
)

type UserRegisteredEvent struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	Timestamp time.Time `json:"timestamp"`
}

type CourseCreatedEvent struct {
	CourseID     string    `json:"course_id"`
	Title        string    `json:"title"`
	InstructorID string    `json:"instructor_id"`
	Timestamp    time.Time `json:"timestamp"`
}

type CoursePublishedEvent struct {
	CourseID  string    `json:"course_id"`
	Title     string    `json:"title"`
	Timestamp time.Time `json:"timestamp"`
}

type EnrollmentStartedEvent struct {
	EnrollmentID string    `json:"enrollment_id"`
	UserID       string    `json:"user_id"`
	CourseID     string    `json:"course_id"`
	Amount       float64   `json:"amount"`
	Timestamp    time.Time `json:"timestamp"`
}

type PaymentProcessedEvent struct {
	PaymentID    string    `json:"payment_id"`
	EnrollmentID string    `json:"enrollment_id"`
	UserID       string    `json:"user_id"`
	CourseID     string    `json:"course_id"`
	Amount       float64   `json:"amount"`
	Status       string    `json:"status"`
	Timestamp    time.Time `json:"timestamp"`
}

type PaymentFailedEvent struct {
	PaymentID    string    `json:"payment_id"`
	EnrollmentID string    `json:"enrollment_id"`
	UserID       string    `json:"user_id"`
	CourseID     string    `json:"course_id"`
	Reason       string    `json:"reason"`
	Timestamp    time.Time `json:"timestamp"`
}

type EnrollmentSuccessEvent struct {
	EnrollmentID string    `json:"enrollment_id"`
	UserID       string    `json:"user_id"`
	CourseID     string    `json:"course_id"`
	Timestamp    time.Time `json:"timestamp"`
}

type EnrollmentFailedEvent struct {
	EnrollmentID string    `json:"enrollment_id"`
	UserID       string    `json:"user_id"`
	CourseID     string    `json:"course_id"`
	Reason       string    `json:"reason"`
	Timestamp    time.Time `json:"timestamp"`
}

type ProgressUpdatedEvent struct {
	UserID             string    `json:"user_id"`
	CourseID           string    `json:"course_id"`
	LessonID           string    `json:"lesson_id"`
	WatchTimeSeconds   int       `json:"watch_time_seconds"`
	ProgressPercentage int       `json:"progress_percentage"`
	Timestamp          time.Time `json:"timestamp"`
}

type LessonCompletedEvent struct {
	UserID    string    `json:"user_id"`
	CourseID  string    `json:"course_id"`
	LessonID  string    `json:"lesson_id"`
	Timestamp time.Time `json:"timestamp"`
}

type CourseCompletedEvent struct {
	UserID    string    `json:"user_id"`
	CourseID  string    `json:"course_id"`
	Timestamp time.Time `json:"timestamp"`
}

type ReviewCreatedEvent struct {
	ReviewID  string    `json:"review_id"`
	UserID    string    `json:"user_id"`
	CourseID  string    `json:"course_id"`
	Rating    int       `json:"rating"`
	Timestamp time.Time `json:"timestamp"`
}
