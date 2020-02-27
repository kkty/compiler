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
let arr = create_array 10 1.0 in if arr.(9) = 1.0 then print_int 1 else print_int 0
