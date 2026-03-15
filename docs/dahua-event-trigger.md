# Dahua Event Trigger to Loxone UDP

Use LoxoneBridge to convert Dahua HTTP event callbacks into UDP messages that Loxone can consume through a Virtual UDP Input.

Replace placeholder values such as `LOXONE`, `PORT`, `loxone-bridge`, and `MyOwnCustomEventName` with values from your environment.

## Dahua device setup

In the Dahua NVR or camera event configuration for the selected trigger:

- Open **More** for the specific event
- Set **Report alarm** = `HTTP`
- Set **Event** = `enabled`
- Set **Send command** =

  ```text
  http://loxone-bridge:8080/udp/LOXONE:PORT/MyOwnCustomEventName
  ```

Replace:
- `LOXONE` with the Loxone Miniserver address
- `PORT` with the UDP port you want to receive notifications on
- `loxone-bridge` with the address of your LoxoneBridge instance
- `MyOwnCustomEventName` with any UDP payload text you want to detect in Loxone

## Loxone setup

1. Add a **Virtual UDP Input**.
2. Set the sender address to the address of LoxoneBridge.
3. Set the port to the chosen UDP port.
4. Add a nested **Virtual UDP Input Command** with:
   - **Use as Digital Input** = `true`
   - **Command Recognition** = `MyOwnCustomEventName`

## Result

You should receive a pulse in Loxone whenever the configured Dahua event is detected.

## Why not directly in Loxone?

This can be done directly, but it typically requires extra workaround steps such as creating a dedicated user, adding a custom button, exposing it in the user interface, and triggering that button through a URL. LoxoneBridge avoids those dummy objects and permissions.
