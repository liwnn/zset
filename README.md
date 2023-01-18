# ZSet
This Go package provides an implementation of sorted set in redis.

## Usage
All you have to do is to implement a comparison `function Less(Item) bool` and a `function Key() string` for your Item which will be store in the zset, here are some examples.
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

func (u User) Key() string {
	return u.Name
}

func (u User) Less(than zset.Item) bool {
	if u.Score == than.(User).Score {
		return u.Name < than.(User).Name
	}
	return u.Score < than.(User).Score
}

func main() {
	zs := zset.New()

	// Add
	zs.Add(User{Name: "Hurst", Score: 88})
	zs.Add(User{Name: "Peek", Score: 100})
	zs.Add(User{Name: "Beaty", Score: 66})

	// Rank
	rank := zs.Rank("Hurst", true)
	fmt.Printf("Hurst's rank is %v\n", rank) // expected 2

	// Range
	fmt.Println()
	fmt.Println("Range[0,3]:")
	zs.Range(0, 3, true, func(v zset.Item, rank int) bool {
		fmt.Printf("%v's rank is %v\n", v.Key(), rank)
		return true
	})

	// Range with Iterator
	fmt.Println()
	fmt.Println("Range[0,3] with Iterator:")
	for it := zs.RangeIterator(0, 3, true); it.Valid(); it.Next() {
		fmt.Printf("Ite: %v's rank is %v\n", it.Item().Key(), it.Rank())
	}

	// Range by score [88, 100]
	fmt.Println()
	fmt.Println("RangeByScore[88,100]:")
	zs.RangeByScore(func(i zset.Item) bool {
		return i.(User).Score >= 88
	}, func(i zset.Item) bool {
		return i.(User).Score <= 100
	}, true, func(i zset.Item, rank int) bool {
		fmt.Printf("%v's score[%v] rank is %v\n", i.Key(), i.(User).Score, rank)
		return true
	})

	// Remove
	zs.Remove("Peek")

	// Rank
	fmt.Println()
	fmt.Println("After remove Peek:")
	rank = zs.Rank("Hurst", true)
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

RangeByScore[88,100]:
Peek's score[100] rank is 1
Hurst's score[88] rank is 2

After remove Peek:
Hurst's rank is 1
```
