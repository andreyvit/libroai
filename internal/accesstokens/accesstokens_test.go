package accesstokens

import (
	"fmt"
	"time"
)

func Example() {
	now := time.Date(2020, 02, 05, 9, 41, 0, 0, time.UTC)
	conf := Configuration{
		Keys:     [][]byte{[]byte("hello world")},
		Prefixes: []string{"TOKEN"},
		Validity: 24 * time.Hour,
	}

	token := conf.SignAt(now, "foo")
	fmt.Println(token)

	t, err := conf.ValidateAt(now, token)
	fmt.Printf("%v %s\n", err, t.DebugString())

	t, err = conf.ValidateAt(now.Add(23*time.Hour), token)
	fmt.Printf("%v %s\n", err, t.DebugString())

	t, err = conf.ValidateAt(now.Add(25*time.Hour), token)
	fmt.Printf("%v %s\n", err, t.DebugString())

	// Output: TOKEN1-20200205094100-foo-19ba44bbedb499922ac0d9df70cdbdb078ba28986171aacf808ed319261e90af
	// <nil> foo from=20200205094100 till=20200206094100
	// <nil> foo from=20200205094100 till=20200206094100
	// expired token <none>
}
