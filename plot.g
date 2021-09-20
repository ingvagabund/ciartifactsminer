set terminal png size 1200, 1200
set output 'KSdegradedmasterupgrade.png'

#set xrange [18:24]
#master upgrade with KS operator degraded
set ylabel "[s]"

plot 'data.dat' using 1:2:3:4:5 with vectors lw 3 lc rgb variable nohead
