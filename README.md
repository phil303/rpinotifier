# RPI Availability Notifier

Scrapes the rpilocator rss feed looking for available Raspberry Pis and sends you a push notification if there is.

## Setup

This uses a push notif backend called [Pushover](https://pushover.net/), which costs a one time fee of $5. 
Register there, adding a custom device. Then set up a new application, which only seems to be doable via the website dashboard.

This will net you a user key, device name, and an application token.

Copy the example-config.yaml file, rename it config.yaml, add the above values into the config and bob's your uncle.

## Running it

The intention for this is to be a cronjob on a server somewhere but I'll leave that as an exercise for the reader. 
To run it, you can either dynamically build and run it or pre-compile and run it.

    go run main.go

or 
    
    go build
    ./rpiscraper