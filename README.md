# Shuffle 'em
It is a library for bit shuffling, written in golang. The usecases are listed further below in the file.

# Installation
`go-shufflem` uses [Go Modules](https://go.dev/blog/using-go-modules). You can use the 
```
go get github.com/vaibhav-kaushal/go-shufflem
```
command (or you can just copy paste the `shuffler.go` file in your codebase if you can't use `go mod`). _The project has no dependencies outside the golang's standard library._

# Testing
## Checking the output 
The `main.go` file can be used to check the output. It contains a sample implementation of the library. You can run 
```
go run main.go shuffler.go
```
from the project root to see it for yourself. 

## Unit testing
To run the unit test, you can run: 

```
go test
```

## Performance testing
The code is not super efficient and the performance will very according to the size of the input and number of shuffles you ask the program to perform. Hence, you should test _your usecase_ for performance. A sample test is already present. To run the benchmark, run:

```
go test -bench=. -count=5 -run=^#
```

Note: The `-run=^#` part requests go to not run the unit tests. Omitting that part should still be okay.

# Usecases
There are multiple places where this library can be useful. 

## 1. Public IDs of objects in a web application 
This library is inspired by possible use of [ULID](https://github.com/ulid/spec). ULID allows you to have decentralized Primary Keys while guaranteeing uniqueness (how you can use ULIDs to guarantee uniqueness is described a little below). At the same time, they keep the index sizes (and thus lookups) faster.

However ULIDs start with a time component and might not be the best for being used at places where the primary keys are not safe for being revealed (such as public URLs) as they automatically indicate the time of creation of an object.

There are two approaches to solve that issue - 

1. Create some kind of `public_id` for each object whose ID can be exposed in the public and then map/search the real ID of the object in your application code.
2. Create a bit-shuffle mapping for each type of object and shuffle the bits before showing them in public URLs and in your application code, reshuffle them before searching for them in the database. 

It is the second approach where this library can come in really handy. You can have one config for each type (e.g. User IDs, Post IDs etc.). Before using them in public (such as URLs), you can do a shuffle. Similarly, when receiving them in your application code from user (such as in an API call), shuffle them before using.

The advantage of this approach is that you never have to save a `public_id`. That allows you to: 
1. Save disk space.
2. Prevent database lookups on `public_id`s.
3. Simplifies queries, especially around joins. 
4. Anyone outside your team working (e.g. someone having the DB dump) on a project who wants to relate any object's Public ID will never find the ID in your database and is going to have a really hard time figuring out how all the IDs in public URLs map to what's in the database.

If you are using a (micro)service based architecture, you can have the shuffler in its own service and conceal the shuffle maps of each type in that service.

### Implementing uniqueness with ULID
If you can use a region or DC-based component in the ULID's randomness bits and incrementing the rest based on an algorithm that only increments the remaining bits within the millisecond. The randomness bits being guaranteed to be unique within the millisecond has already been implemented well in golang ([oklog's UUID library](https://github.com/oklog/ulid)). You can use the same library to define your own entropy for the above-mentioned behavior.

## 2. Custom symmetric-encryption-like behavior
Most symmetric encryption algorithms are based on three core parts: 

1. **Blocks of data**: Each encryption mechanism ingests data in blocks of predefined size.
2. **Encryption algorithm**: The core algorithm which changes the input blocks to the encrypted (output) blocks.
3. **An encryption key**: A set of bytes which change the way algorithm will encrypt the blocks. Changing the key while keeping the algorithm and the input will result in a different output in almost all encryption algorithms.

You can use this library to encrypt your data (in a way) where the shuffle map will serve as the encryption key. Like _real_ encryption algorithms, when using bit-shuffling for encryption, guessing the key or original data becomes more and more difficult as you increase the `BitCount` and number of shuffle pairs in `Config`!

The advantage here is that you can use can vary the input block size (increase the `BitCount`; ensure it is a multiple of 8 though) and strength of encryption (number of entries in `ShuffleMap`) according to your choice!

For example, if the input (in hexadecimal notation)`676f73687566666c656d6c696240766169626861766b61757368616c2e636f6d` gets changed to `b6f6c62e3686166bae86d66e0d164696866e021a9636b6a6746666ae68cef6e6`, can you really tell what the shuffle map was like? How long would it take for you to make the correct guess and how many such samples are you going to need for the correct guess?

# Todo
1. Add the cyclic shuffling capability.
2. Add an example in this (README) file.

## What is bit swapping
Bit swapping is swapping the value of bits. Assuming your input bits (spaces are added after 8 bits for readability only) are `11001101 10101101` and you want to swap bits at indexes (indexed from `0`) `1` and `11` and bits `7` and `8` then your result will be `10001110 10111101`

## What is meant by cyclic shuffling?
When you interchange the positions of a set of bits in a way that they are not simple shuffles, but repositioned amongst themselves, it would be called cyclic shuffling.

Assuming the same input again (`11001101 10101101`) if you want to set value of bit `1` to value at bit `6`, value of bit `6` to value at bit `8` and value of bit `8` to value at bit `1`, then your result would be `10001111 11101101`. This looks like a simple bit swap too but with various inputs, the results will vary in ways that will be much more difficult to predict what the shuffle map was.



