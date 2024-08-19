package models

type Parent struct {
	AssociatedAccount  string   `firestore:"associated_account"`
	AssociatedStudents []string `firestore:"associated_students"`
}
