package packagex

type User struct {
	Name string
	Age  int
}

type Book struct {
	Name    string
	Content struct {
		Text string
	}
}

func (u User) Buy(b *Book) error {
	return nil
}

var user User
var book Book

func Do() {
	err := user.Buy(&book)
	if err != nil {
		panic(err)
	}
}
