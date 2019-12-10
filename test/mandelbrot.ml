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
let rec itof x =
  let rec f x y = if x = 0 then y else f (x - 1) (y +. 1.0) in
  f x 0.0
in
let rec dbl f = f +. f in
let rec yloop y =
  if y >= 40 then ()
  else
    let rec xloop x y =
      if x >= 40 then ()
      else
        let cr = (dbl (itof x) /. 40.0) -. 1.5 in
        let ci = (dbl (itof y) /. 40.0) -. 1.0 in
        let rec iloop i zr zi zr2 zi2 =
          if i = 0 then print_int 1
          else
            let tr = zr2 -. zi2 +. cr in
            let ti = (dbl zr *. zi) +. ci in
            let zr = tr in
            let zi = ti in
            let zr2 = zr *. zr in
            let zi2 = zi *. zi in
            if zr2 +. zi2 > 2.0 *. 2.0 then print_int 0
            else iloop (i - 1) zr zi zr2 zi2
        in
        iloop 10 0.0 0.0 0.0 0.0 ;
        xloop (x + 1) y
    in
    xloop 0 y ;
    yloop (y + 1)
in
yloop 0
