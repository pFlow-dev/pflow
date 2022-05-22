local function pos(x, y)
	return { x=x*100, y=y*80, z=0 }
end

domodel("collection", function (fn, cell, role)

	local default = role("user")
	local owner = role("owner")
	local author = role("author")
	local admin = role("admin")

	function Pages()
		local pages = cell("page.count", 0, 0, pos(4,6))
		rm = fn("page.rm", owner, pos(5,1))
		pages.tx(1, rm)

		add = fn("page.add", author, pos(3,1))
		add.tx(1, pages )

		update = fn("page.update", owner, pos(4,1))
		view = fn("page.view", default, pos(6, 1))

		return {
			pages = pages,
			view = view,
			fn = {
				add = add,
				update = update,
				rm = rm,
			}
		}
	end

	function User()
		signup = fn("user.signup", default, pos(1, 5))
		online = fn("user.online", default, pos(1, 2))
		login = fn("user.login", default, pos(1, 4))
		logout = fn("user.logout", default, pos(1, 1))
		reset = fn("user.reset", default, pos(1, 3))

		exist = cell("user.exist", 1, 1, pos(1, 6))
		exist.tx(1, signup)
		exist.guard(1, online)
		exist.guard(1, login)
		exist.guard(1, logout)
		exist.guard(1, reset)

		offline = cell("user.offline", 1, 1, pos(6, 6))
		offline.guard(1, online)
		offline.tx(1, login)
		logout.tx(1, offline)
		reset.tx(1, offline)

		return {
			offline = offline,
			fn = {
				online = online,
				offline = offline,
				login = login,
				logout = logout
			}
		}
	end

	function Collection()
		exist = cell("collection.exist", 1,1, pos(3,6))
		disable = fn("collection.disable", owner, pos(2,3))
		disable.tx(1, exist)
		create = fn("collection.create", author, pos(2,1))
		exist.tx(1, create)
		publish = fn("collection.publish", author, pos(2,2))
		tag = cell("collection.tag", 0,0, pos(5,6))
		publish.tx(1, tag)

		return {
			exist = exist,
			create = create,
			fn = {
				disable = disable,
				publish = publish
			}
		}
	end

    version = cell("version", 0, 0, pos(2,6))
    user = User()
    collection = Collection()
    user.offline.guard(1, collection.create)

    function version_edits_and_assert_online(n)
        n.tx(1, version)
        user.offline.guard(1, n)
        collection.exist.guard(1, n)
    end

    for _, n in pairs( collection.fn ) do
        version_edits_and_assert_online(n)
    end

    pages = Pages()
    collection.exist.guard(1, pages.view)
    for _, n in pairs( pages.fn ) do
        version_edits_and_assert_online(n)
    end

end)
