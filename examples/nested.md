#!(var colors = red|blue|green)
#!(var names = dave|john|bob)
#!(
	foreach name in #!(var names) {
		#!(foreach color in #!(var colors) {
			#!(var name) - #!(var color)\n
		})
	}
)

arg 1: #$0
arg 2: #$1
arg 3: #$2

args: #$...
