#### Files:
#!(foreach file in #!(list #!(exec ls #$1)) {
`- #!(var file)`\n
})