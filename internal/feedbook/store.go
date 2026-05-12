package feedbook

type Storer interface {
	AvatarPresets() []AvatarPreset
	OwnProfile() Profile
	SetOwnProfile(Profile)
	PublicProfile() Profile
	Stats() Stats
	Notifications() Notifications
	Books() []Book
	BookByID(string) (Book, bool)
	ExploreUsers() []ExploreUser
	ReadingProgress(string) (ReadingProgress, bool)
	SetReadingProgress(bookID string, currentPage int, totalPages int) error
	Reviews(string) []Review
	SaveReview(bookID string, review Review) error
	Authors() []Author
	AuthorByID(string) (Author, bool)
	ToggleFollow(string) bool
	AddBookToLibrary(bookID string) error
	RemoveBookFromLibrary(bookID string) error
}
