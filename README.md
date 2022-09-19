# ZSet
This Go package provides an implementation of sorted set in redis.

## Usage
All you have to do is to implement a comparison `function Less(Item) bool` for your Item which will be store in the zset, here are some examples.
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
	fmt.Println()
	fmt.Println("Range[0,3]:")
	tr.Range(0, 3, true, func(key string, v zset.Item, rank int) bool {
		fmt.Printf("%v's rank is %v\n", key, rank)
		return true
	})

	// Range with Iterator
	fmt.Println()
	fmt.Println("Range[0,3] with Iterator:")
	for it := tr.RangeIterator(0, 3, true); it.Valid(); it.Next() {
		fmt.Printf("Ite: %v's rank is %v\n", it.Key(), it.Rank())
	}

	// Remove
	tr.Remove("Peek")

	// Rank
	fmt.Println()
	fmt.Println("After remove Peek:")
	rank = tr.Rank("Hurst", true)
	fmt.Printf("Hurst's rank is %v\n", rank) // expected 1
}
```
Output:
```
Hurst's rank is 2

Range[0,3]:
Peek's rank is 1
Hurst's rank is 2
Beaty's rank is 3

Range[0,3] with Iterator:
Ite: Peek's rank is 1
Ite: Hurst's rank is 2
Ite: Beaty's rank is 3

After remove Peek:
Hurst's rank is 1
```