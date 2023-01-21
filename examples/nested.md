#!(var colors = red|blue|green)
#!(var names = dave|john|bob)
#!(
	foreach name in #!(var names) {
		#!(foreach color in #!(var colors) {
			#!(var name) - #!(var color)\n
		})
	}
)
