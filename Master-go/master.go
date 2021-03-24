package master

import (
  "Network-go"
  "ftm"
)

/*
Master-modulen får inn ordre og elevator states.
Regner ut ny fordeling av ordre og sender denne ut.
*/

for{
  select{
    /*
    hvis master:
    du får en updated state og updated order inn på channel
    når du får state/order, regner du ut hva ny ordreliste skal være og sender ut til alle og lagrer ny state og order
    */
  }
}

func calculate_distribution(states, orders) {

}
