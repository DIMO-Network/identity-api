# h**ID**e  [![Build Status](https://api.travis-ci.org/c2h5oh/hide.svg?branch=master)](https://travis-ci.org/c2h5oh/hide)  [![GoDoc](https://godoc.org/github.com/c2h5oh/hide?status.svg)](https://godoc.org/github.com/c2h5oh/hide)  [![Go Report Card](https://goreportcard.com/badge/github.com/c2h5oh/hide)](https://goreportcard.com/report/github.com/c2h5oh/hide)

Super easy ID obfuscation that actually works

## The why
IDs autoassigned by database (autoincrement, sequence, etc) leak quite a bit of information about your data:

* the highest ID leaks how many rows (users, products, etc) have been created
* the change of the highest ID over time leaks how fast rows (users, products, etc) are being added, how fast you grow
* the change in speed of the highest ID increase over time leaks if and how fast your growth is accelerating
* iterating over IDs makes scraping all items easy. You can often extract a lot from that too, ie:
  * how many are inactive/deleted and how has that changed over time (churn)
  * the actual content
* and more

At the same time autoincrement IDs are great: they maintain sort order, are dense (more efficient, smaller indexes), can be efficiently stored in DB, guaranteed to be unique and provide you the same information they leak at glance...


## The how

### Rejected solutions:

* random IDs:
  * order is not maintained
  * sparse - larger less efficient indexes, harder partitioning/sharding
  * as you get more rows you will randomly get an existing ID and will have to retry: increased complexity, non-deterministic insert time
* UUIDs:
  * order is not maintained
  * sparse - see above
  * looooooong
  * not the best pick for primary keys in relational databases - less efficient than integers, especially on joins, significantly larger (storage, indexes)
* Time-based IDs:
  * leak most of the information that autoincrement IDs leak - it just requires a little extra effort
  * in addition to that leak exactly when a row has been created
  * less dense than autoincrement IDs
  * all hell breaks loose when system clock is adjusted
  * require extra logic to avoid duplicates if 2 items are inserted within the smallest unit of time used
* Base64-encoded IDs (Youtube-like-looking - ID base64 encoding is not all Youtube is doing):
  * they are just regular autoincrement IDs displayed in a different format that is trivial to reverse


### Goal

* well obfuscated - hard to figure out from outside even if you know the method
* integer IDs, at least in the database. Preferably still consecutive or at least not sparse
* maintained order (older ID < newer ID), at least in database
* as little overhead as possible
* as little changes to existing code as possible


### Solution

Three words: `Modular multiplicative inverse`. Math warning: https://en.wikipedia.org/wiki/Modular_multiplicative_inverse

* To obfuscate ID calculate MMI of the ID using a large prime number
* To deobfuscate ID calculate MMI of the obfuscated ID using inverse of the previously used prime
* You can still use autoincrement integers as IDs internally, with all of the benefits
* Obfuscated IDs look random
* Figuring which prime was used is not easy and brute-forcing it will be hard - even for `int32` there are close to **200.000.000** primes to choose from. You can also set a value used to XOR obfuscated IDs making brute-forcing the prime A LOT more difficult
* Performance is great - this implementation uses highly optimized functions used by Go `crypto/*` packages


# Usage example

Before:
```go
type User struct {
    ID int64 `db:"id" json:"id"`
    Username string `db:"username" json:"username"`
}
```

After:
```go
import "github.com/c2h5oh/hide"

type User struct {
    ID hide.Int64 `db:"id" json:"id"`
    Username string `db:"username" json:"username"`
}
```
That's it. Really. ID will be transparently obfuscated on `json.Marshal` and deobfuscated on `json.Unmarshal`

**(∩｀-´)⊃━☆ﾟ.*･｡ﾟ**

Also supported: `int32`, `uint32`, `int64`, `uint64`

## Sample Results
```
Random IDs
   69407 -> 1679933185
  365732 -> 1149554396
  490883 -> 1588788253
   20826 ->  342781798
  196984 ->   79257480
  849265 -> 1757383279
  235515 -> 1521361573
  649322 ->  694474326
  869519 ->  688585617
  236378 -> 1477305702

Consecutive IDs
  308035 -> 1178570333
  308036 ->  531536956
  308037 -> 2031987227
  308038 -> 1384953850
  308039 ->  737920473
  308040 ->   90887096
  308041 -> 1591337367
  308042 ->  944303990
  308043 ->  297270613
  308044 -> 1797720884
```
Run the example in `cmd` for more.


# Remember to set your own primes!
Package comes with default ~~primes set, but~~ primes NOT SET please pick your own. Good source: https://primes.utm.edu/lists/small/small.html
```go
hide.Default.SetInt32(myInt32Prime)   // set prime used for int32 obfuscation
hide.Default.SetUint32(myUint32Prime) // set prime used for uint32 obfuscation
hide.Default.SetInt64(myInt64Prime)   // set prime used for int64 obfuscation
hide.Default.SetUint64(myUint64Prime) // set prime used for uint64 obfuscation
```

Optionally (highly recommended) set a value used to XOR the resulting ID making discovering the prime used for obfuscation way more difficult
```go
hide.Default.SetXor(myXor)
```

See `obfuscation_test.go` if you need an example.


# Benchmarks
on i7 6700K running Ubuntu 15.10 and go1.6
```
BenchmarkInt32Obfuscate-8 	10000000	       110 ns/op	      48 B/op	       1 allocs/op
BenchmarkInt64Obfuscate-8 	10000000	       109 ns/op	      48 B/op	       1 allocs/op
BenchmarkUint32Obfuscate-8	10000000	       108 ns/op	      48 B/op	       1 allocs/op
BenchmarkUint64Obfuscate-8	10000000	       110 ns/op	      48 B/op	       1 allocs/op
```

# Limitations
The purpose of this package is to obfuscate IDs. **Obfuscate and NOT encrypt**. The transformation used is simple and fast, which makes it a great option for a low-overhead obfuscation, but the same reasons make it brute-forceable should anyone be determined enough. Using XOR makes things better, provided brute-force is used - I would be amazed if multiple attacks reducing the complexity didn't exist.
