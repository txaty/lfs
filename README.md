# Lagrange Four Square Sum

Lagrange Four Square Sum is a Go package that implements an algorithm to solve the Lagrange four-square sum problem for large integers. 
The algorithm computes a representation of a positive integer `n` as a sum of four squares:

$$
n = {w_0}^2 + {w_1}^2 + {w_2}^2 + {w_3}^2
$$

This implementation is highly optimized for very large integers (e.g., 1000 bits, 2000 bits, or more).

The algorithms are based on paper, [Finding the Four Squares in lagrange's Theorem](https://campus.lakeforest.edu/trevino/finding4squares.pdf) with improvements that significantly improve the speed.

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

- [txaty/go-bigcomplex](https://github.com/txaty/go-bigcomplex): A big complex number library supporting big Gaussian and Hurwitz integers.
- [lukechampine.com/frand](https://github.com/lukechampine/frand): A fast randomness generation library.

## License

This project is released under the MIT License.