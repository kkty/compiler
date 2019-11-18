let rec fib n = if n < 2 then 1 else fib (n - 1) + fib (n - 2) in
let i = fib 10 in
print_int i
