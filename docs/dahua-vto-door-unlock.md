# Dahua VTO Door Unlock

Use LoxoneBridge to translate Loxone Basic Authentication into upstream Digest Authentication for a Dahua intercom door unlock command.

Replace placeholder values such as `DAHUA_USER`, `DAHUA_PASSWORD`, `INTERCOM_ADDRESS`, and `loxone-bridge` with values from your environment.

## Loxone setup

1. Create a **Virtual Output** in Loxone Miniserver with this address:

   ```text
   http://USER:PASSWORD@loxone-bridge:8080/
   ```

   Replace `DAHUA_USER` and `DAHUA_PASSWORD` with the Dahua login.

2. Create a nested **Virtual Output Command** with:
   - **Use as Digital Output** = `true`
   - **Command for On** =

     ```text
     /digest/http/INTERCOM_ADDRESS/cgi-bin/accessControl.cgi?action=openDoor&channel=1
     ```

   - **HTTP Method for On** = `GET`

   Replace `INTERCOM_ADDRESS` with the real intercom IP address or hostname.

## Result

When this virtual output is triggered, the door should unlock.

## Why not directly in Loxone?

Loxone does not support upstream HTTP Digest Authentication, while Dahua expects it for this endpoint.
