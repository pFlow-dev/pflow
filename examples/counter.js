
function pos(x, y) {
    return { x: x*100, y: y*80, z: 0 }
}

function CounterModel (fn, cell, role) {
    r = role("default")

    p1 = cell("p1", 0, 1, pos(1, 1))
    p2 = cell("p2", 0, 0, pos(1,2))
    p3 = cell("p3", 1, 1, pos(1, 3))

    inc1 = fn("inc1", r, pos(2, 1))
    inc1.tx(1, p1)

    dec1 = fn("dec1", r, pos(2, 2))
    p1.tx(1, dec1)

    inc2 = fn("inc2", r, pos(3, 1))
    inc2.tx(1, p2)

    dec2 = fn("dec2", r, pos(3, 2))
    p2.tx(1, dec2)

    inc3 = fn("inc3", r, pos(4, 1))
    inc3.tx(1, p3)

    dec3 = fn("dec3", r, pos(4, 2))
    p3.tx(1, dec3)

    p3.guard(1, inc1)
}

domodel("counter-js", CounterModel)
