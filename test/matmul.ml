let rec print_int x =
  let y = x >= 100 in
  let x =
    if x >= 200 then (print_char 50; x - 200)
    else if x >= 100 then (print_char 49; x - 100)
    else x in
  let x =
    if x >= 90 then (print_char 57; x - 90)
    else if x >= 80 then (print_char 56; x - 80)
    else if x >= 70 then (print_char 55; x - 70)
    else if x >= 60 then (print_char 54; x - 60)
    else if x >= 50 then (print_char 53; x - 50)
    else if x >= 40 then (print_char 52; x - 40)
    else if x >= 30 then (print_char 51; x - 30)
    else if x >= 20 then (print_char 50; x - 20)
    else if x >= 10 then (print_char 49; x - 10)
    else ((if y then print_char 48 else ()); x) in
  if x = 9 then print_char 57
  else if x = 8 then print_char 56
  else if x = 7 then print_char 55
  else if x = 6 then print_char 54
  else if x = 5 then print_char 53
  else if x = 4 then print_char 52
  else if x = 3 then print_char 51
  else if x = 2 then print_char 50
  else if x = 1 then print_char 49
  else print_char 48 in
let rec ftoi x =
  let rec f x y = if x < 0.00001 then y else f (x -. 1.0) (y + 1) in
  f x 0
in
let rec mul l m n a b c =
  let rec loop1 i =
    if i < 0 then ()
    else
      let rec loop2 j =
        if j < 0 then ()
        else
          let rec loop3 k =
            if k < 0 then ()
            else (
              c.(i).(j) <- c.(i).(j) +. (a.(i).(k) *. b.(k).(j)) ;
              loop3 (k - 1) )
          in
          loop3 (m - 1) ;
          loop2 (j - 1)
      in
      loop2 (n - 1) ;
      loop1 (i - 1)
  in
  loop1 (l - 1)
in
let dummy = create_array 0 0. in
let rec make m n =
  let mat = create_array m dummy in
  let rec init i =
    if i < 0 then ()
    else (
      mat.(i) <- create_array n 0. ;
      init (i - 1) )
  in
  init (m - 1) ;
  mat
in
let a = make 2 3 in
let b = make 3 2 in
let c = make 2 2 in
a.(0).(0) <- 1. ;
a.(0).(1) <- 2. ;
a.(0).(2) <- 3. ;
a.(1).(0) <- 4. ;
a.(1).(1) <- 5. ;
a.(1).(2) <- 6. ;
b.(0).(0) <- 7. ;
b.(0).(1) <- 8. ;
b.(1).(0) <- 9. ;
b.(1).(1) <- 10. ;
b.(2).(0) <- 11. ;
b.(2).(1) <- 12. ;
mul 2 3 2 a b c ;
print_int (ftoi c.(0).(0)) ;
print_int (ftoi c.(0).(1)) ;
print_int (ftoi c.(1).(0)) ;
print_int (ftoi c.(1).(1))
