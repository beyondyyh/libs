package set

import (
	"testing"
)

type Guy struct {
	Name    string
	Age     int
	Company string
}

// run: go test -v -run Test_ExampleIterator
func Test_ExampleIterator(t *testing.T) {
	set := NewSetWithSlice([]interface{}{
		&Guy{Name: "John", Age: 21, Company: "Yahoo"},
		&Guy{Name: "Yehong.Yang", Age: 18, Company: "Weibo"},
		&Guy{Name: "Bob", Age: 22, Company: "Aalibaba"},
		&Guy{Name: "Nick", Age: 23, Company: "Tencent"},
		&Guy{Name: "Mary", Age: 17, Company: "Bytedance"},
	})

	var handsomest *Guy
	it := set.Iterator()

	for ele := range it.C {
		if ele.(*Guy).Name == "Yehong.Yang" {
			handsomest = ele.(*Guy)
			it.Stop()
		}
	}

	t.Logf("Handsomest guy is %+v\n", handsomest)
}
