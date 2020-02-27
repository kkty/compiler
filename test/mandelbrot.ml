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
