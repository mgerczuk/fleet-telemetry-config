# Tesla Fleet Telemetry Configuration
--------------------------------

This tool helps with sending a Fleet Telemetry configuration to one or more vehicles. It starts with creating and registering an application at the Tesla developer site.

You can set one or more users with their Tesla accounts to retrieve tokens and pair the application key to their vehicles. Finally it allows to send a telemetry configuration to the vehicles.

For this purpose it creates two HTTP servers:

* A public website that should run on port 443 which handles everything needed for interaction with Tesla
* A private website which offers a UI for getting Tesla access tokens, pair a key with the car and send a telemetry configuration to the car(s)

## Public Website

**Important: This website must be accessible from the public internet with the URL entered when registering the application at the Tesla developer site!**

The following URLs are handled

`/.well-known/appspecific/com.tesla.3p.public-key.pem`

This URL is used by the Tesla server to query the public key.

`/auth/callback`

This URL is used by the Tesla server to forward the initial token to this tool.

`/robots.txt`

This hopefully reduces traffic from search engines and other crawlers.

## Private Website

**Important: This website should not be accessible from the public internet!**

If the server is located in your LAN then do not forward the port in your router. If it is located outside your LAN (e.g. VPS) you may create an SSH tunnel.

On the main page of the Website you should first follow the steps to setup the application.

After this is done you can create at least one user and press the "Token" button to get the access token and pair the key to a vehice. With the token you can press the "Configure" button to send a Fleet Telemetry configuration to the vehicle(s).
