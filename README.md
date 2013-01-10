# razfaz

This is the game plan for the volleyball team [Raz Faz](http://tvwollishofen.ch/volley/teams). 

It is a clone of the website done by Iwan Birrer, implemented in go (his original website is javascript/nodejs). At the time of writing, it is available at [http://razfaz.there.ch](http://razfaz.there.ch)

The code parses the [RVZ](http://r-v-z.ch) website for the game plan of the team and show it as a nice webpage.

## Requirements

An installation of [Go](http://golang.org) that supports html parsing in `exp/html`. At the time of writing, this is available when installing Go from source and using `hg tip`. It will likely be available in Go 1.1 in package `html`.


## Installation

    git clone https://github.com/eglimi/razfaz.git
    cd razfaz
    go build razfaz
    ./razfaz

Point your browser to [http://localhost:8080](http://localhost:8080).
