# spawn
 a fleetd wrangler for coreos
 Simon Kirkby
 tigger@interthingy.com
 20150521

 # Usage

 Source and target IP addresses can be set with flags or environmental variables

 ## Source 

 the spawn web service 

 export SPWAN_SOURCE=10.10.10.5 , or -source=10.10.10.5

 ## Target 

 the fleetd machines

 export SPWAN_SOURCE=10.10.10.5 , or -source=10.10.10.5

 # Function 

 spawn connects astralboot to fleetd and wrangles unit files and .aci files 

 # TODO 

1. manage list of running units
2. generate unit files
3. generate sidekicks
4. start and stop units according to authoritive list
5. More stuff
