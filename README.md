# LFS - Lagrange Four-Squares

[![Go Reference](https://pkg.go.dev/badge/github.com/txaty/lfs.svg)](https://pkg.go.dev/github.com/txaty/lfs)
[![Go Report Card](https://goreportcard.com/badge/github.com/txaty/lfs)](https://goreportcard.com/report/github.com/txaty/lfs)
[![codecov](https://codecov.io/github/txaty/lfs/graph/badge.svg?token=ggYw40inA4)](https://codecov.io/github/txaty/lfs)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/fa8bf34169a242d58a1d952988f2e81e)](https://app.codacy.com/gh/txaty/lfs/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)

LFS is a Go package that implements an optimized algorithm for solving the Lagrange four-squares problem, even for very large integers.
It computes a representation of any positive integer `n` as the sum of four squares:

$$
n = {w_0}^2 + {w_1}^2 + {w_2}^2 + {w_3}^2
$$

This implementation is highly optimized for very large integers (e.g., 1000 bits, 2000 bits, or more).

The underlying algorithms are based on Section 3 of the paper [Finding the Four Squares in Lagrange's Theorem](https://campus.lakeforest.edu/trevino/finding4squares.pdf) with additional improvements that significantly enhance performance.

## Usage

Below is a simple example:

```go
package main

import (
    "crypto/rand"
    "fmt"
    "log"
    "math/big"

    "github.com/txaty/lfs"
)

func main() { 
    // Create a new solver with default options.
    solver := lfs.NewSolver()

    // Define a large integer.
    limit := new(big.Int).Lsh(big.NewInt(1), 1200)
    n, err := rand.Int(rand.Reader, limit)
    if err != nil {
        log.Fatal(err)
    }

    // Compute the four-square representation.
    result := solver.Solve(n)

    // Display the result.
    fmt.Printf("Representation of n as sum of four squares: %s\n", result)

    // Verify the representation.
    if lfs.Verify(n, result) {
        fmt.Println("Verification succeeded: The squares sum to n.")
    } else {
        log.Fatal("Verification failed: The computed squares do not sum to n.")
    }
}
```

## Configuration Options

The solver is configurable via functional options when creating a new instance. For example:

- **WithFCMThreshold**: Sets the threshold above which the advanced FCM algorithm is used.
  Example:
    ```go
    solver := lfs.NewSolver(
        lfs.WithFCMThreshold(new(big.Int).Lsh(big1, 600)), // Use FCM for numbers ≥ 2^600
    )
    ```
- **WithNumRoutines**: Sets the number of concurrent goroutines for the randomized search.
  Example:
    ```go
    solver := lfs.NewSolver(
        lfs.WithNumRoutines(8), // Use 8 goroutines for parallel computation
    )
    ```

## Dependencies

This project requires the following dependencies:

- [txaty/go-bigcomplex](https://github.com/txaty/go-bigcomplex): A big complex number library supporting big Gaussian
  and Hurwitz integers.
- [lukechampine.com/frand](https://github.com/lukechampine/frand): A fast randomness generation library.

## License

This project is released under the MIT License.