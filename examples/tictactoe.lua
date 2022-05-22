domodel("TicTacToe", function (fn, cell, role)

	local dy = 140
	local dx = 220

	local function row(n)
	    return {
	    	[0] = cell(n..0, 1, 1, { x=1*dx, y=(n+1)*dy }),
	    	[1] = cell(n..1, 1, 1, { x=2*dx, y=(n+1)*dy }),
	    	[2] = cell(n..2, 1, 1, { x=3*dx, y=(n+1)*dy })
	    }
	end

	local board = {
		[0] = row(0),
		[1] = row(1),
		[2] = row(2)
	}

	local X, O = "X", "O"

	local players = {
		[X] = {
			turn = cell(X, 1, 1, { x=40, y=200 }), -- track turns, X goes first
			role = role(X), -- player X can only mark X's
			next = O, -- O moves next
			dx = -60 -- position moves to the left of cell
		},
		[O] = {
			turn = cell(O, 0, 1, { x=830, y=370 }), -- track turns, moves second
			role = role(O), -- player O can only mark O's
			next = X, -- X moves next
			dx = 60 -- position moves to the right of cell
		}
	}

	for i, board_row in pairs(board) do
		for j in pairs(board_row) do
			for marking, player in pairs(players) do
				local pos = board_row[j].place.position
				local move = fn(marking..i..j, player.role, { -- make a move
					x=pos.x+player.dx,
					y=pos.y
				})
				player.turn.tx(1, move) -- take turn
				board[i][j].tx(1, move) -- take board space
				move.tx(1, players[player.next].turn) -- mark next turn
			end
		end
	end

end)
