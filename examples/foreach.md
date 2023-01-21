#!(var a = 5)
I have #!(var a) apples!

#!(
	foreach name in "john steve billy" {
		`Hey #!(var name)! `
	}
)

# foreach loops:

#!(var names = "100 200 300 400 500")
#!(
	foreach name in #!(list #!(var names)) {
		![](https://via.placeholder.com/#!(var name)x#!(var name).png)\n
	}
)


#!(
	foreach name in #!(list #!(var names)) {
		#!(exec cowsay #!(var name))
	}
)

# List Files



#!(var dir = ".")
Files in #!(exec pwd):
#!(
	foreach file in 
	#!(trim 1 -3 #!(list #!(exec ls -l #!(var dir) ) ) )
	{
		`- #!(var file)`\n
	}
)
