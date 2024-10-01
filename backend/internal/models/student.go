package models

type Student struct {
	AssociatedAccount string   `firestore:"associated_account"`
	AssociatedTutors  []string `firestore:"associated_tutors"`
	Curriculum        []string `firestore:"curriculum"`
	Email             string   `firestore:"email"`
	Name              string   `firestore:"name"`
	LifetimeHours     float64  `firestore:"lifetime_hours"`
	RemainingHours    float64  `firestore:"remaining_hours"`
	UsedHours         float64  `firestore:"used_hours"`
}
