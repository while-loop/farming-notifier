# Farming notifier

Send an SMS via Twilio when a farming patch is ready for harvesting in Oldschool RuneScape.


Farming info is supplied via a custom [RuneLite](https://github.com/runelite/runelite) Farming plugin.

The development cycle
---------------------

* Make changes to the .go files
* Run `make` to compile the binaries to Linux AMD64
* Run `serverless deploy` to construct the service and push changes to lambda
