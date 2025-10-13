package main

func fib(n int) int {
	if n == 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	return fib(n-1) + fib(n-2)
}

func OptimizedFib(n int) int {
	if n == 0 {
		return 0
	}
	a := 0
	b := 1
	for i := 2; i <= n; i++ {
		tmp := a + b
		a = b
		b = tmp
	}
	return b
}
