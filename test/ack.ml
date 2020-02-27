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
let rec ack x y =
  if x <= 0 then y + 1
  else if y <= 0 then ack (x - 1) 1
  else ack (x - 1) (ack x (y - 1))
in
print_int (ack 3 5)
