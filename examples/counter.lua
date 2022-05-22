require "metamodel"

local function pos(x, y)
	return { x=x*100, y=y*80, z=0 }
end

domodel("counter", function (fn, cell, role)
	local user = role("user")

	local p0 = cell('p0', 1, 0, pos(2, 1))
	local inc0 = fn('inc0', user, pos(1, 1))
	inc0.tx(1, p0)
	local dec0 = fn('dec0', user, pos(3, 1))
	p0.tx(1, dec0)

	local p1 = cell('p1', 0, 1, pos(2, 2))
	local inc1 = fn('inc1', user, pos(1, 2))
	inc1.tx(1, p1)
	local dec1 = fn('dec0', user, pos(3, 2))
	p1.tx(1, dec1)

	local p2 = cell('p2', 0, 1, pos(2, 3))
	p2.guard(1, inc0)
	local inc2 = fn('inc2', user, pos(1, 3))
	inc2.tx(1, p2)
	dec2 = fn('dec2', user, pos(3, 3))
	p2.tx(1, dec2)
end)
