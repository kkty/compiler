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
let rec gcd m n =
  if m = 0 then n else if m <= n then gcd m (n - m) else gcd n (m - n)
in
print_int (gcd 72 120)
