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

type BookWithTag struct {
	Name   string `json:"name,omitempty" gorm:"column:book;type:varchar(255);"`
	Author string `json:"author,omitempty" gorm:"column:author;type:varchar(30);"`
}

type BookWithTypeParam[T any] struct {
	Name  string
	Value T
}
