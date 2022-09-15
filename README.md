# ZSet
This Go package provides an implementation of sorted set in redis.

## Usage
``` go
package main

import (
	"fmt"

	"github.com/liwnn/zset"
)

type User struct {
	Name  string
	Score int
}

func (u User) Less(than zset.Item) bool {
	if u.Score == than.(User).Score {
		return u.Name < than.(User).Name
	}
	return u.Score < than.(User).Score
}

func main() {
	tr := zset.New()

	// Add
	tr.Add("Hurst", User{Name: "Hurst", Score: 88})
	tr.Add("Peek", User{Name: "Peek", Score: 100})
	tr.Add("Beaty", User{Name: "Beaty", Score: 66})

	// Rank
	rank := tr.Rank("Hurst", true)
	fmt.Printf("Hurst's rank is %v\n", rank) // expected 2

	// Range
	rg := tr.Range(0, 3, true)
	for i, v := range rg {
		fmt.Printf("%v's rank is %v\n", v.(User).Name, i+1)
	}

	// Delete
	tr.Delete("Peek")

	// Rank
	rank = tr.Rank("Hurst", true)
	fmt.Printf("Hurst's rank is %v\n", rank) // expected 1
}
```
Output:
```
Hurst's rank is 2
Peek's rank is 1
Hurst's rank is 2
Beaty's rank is 3
Hurst's rank is 1
```