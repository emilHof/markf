#!(foreach name in #!(trim 1 -1 #$...) {
	Hello #!(var name)!\n
})