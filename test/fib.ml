let rec print_int x =
  let x =
    if x >= 100 then
      print_char (48 + x / 100);
      x - (x / 100) * 100
    else x in
  let x =
    if x >= 10 then
      print_char (48 + x / 10);
      x - (x / 10) * 10
    else x in
  print_char (48 + x) in
let rec fib n = if n < 2 then 1 else fib (n - 1) + fib (n - 2) in
let i = fib 10 in
print_int i
