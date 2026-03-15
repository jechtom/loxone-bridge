# Shelly Communication via Digest and Flattened JSON

Use LoxoneBridge to authenticate against a Shelly device with Digest Authentication and flatten the JSON response into a `key=value` structure that is easier to parse in Loxone.

Replace placeholder values such as `PASSWORD`, `SHELLY-ADDRESS`, and `loxone-bridge` with values from your environment.

## Loxone setup

1. Add a **Virtual HTTP Input** in Loxone Miniserver.
2. Use this address:

   ```text
   http://admin:PASSWORD@loxone-bridge:8080/digest/flatten-json/http/SHELLY-ADDRESS/rpc/Shelly.GetStatus
   ```

   Replace:
   - `PASSWORD` with the Shelly password
   - `SHELLY-ADDRESS` with the Shelly device address
   - `loxone-bridge` with your LoxoneBridge instance address if needed

3. Add a nested **Virtual HTTP Input Command** for the value you want to read.
   - Example **Command Recognition**:

     ```text
     aenergy.total
     ```

## Result

Loxone receives a flat response where nested JSON values can be addressed by a stable key path.

## Why not directly in Loxone?

Digest Authentication is not implemented in Loxone for this type of request, and parsing nested JSON with repeated field names can be difficult and fragile with command recognition patterns.

## Original Shelly response

```json
{"id":0, "source":"init", "output":true, "apower":193.5, "voltage":242.7, "freq":50.0, "current":0.883, "aenergy":{"total":52648.590,"by_minute":[3403.476,3190.759,3190.759],"minute_ts":1773617820}, "ret_aenergy":{"total":0.000,"by_minute":[0.000,0.000,0.000],"minute_ts":1773617820},"temperature":{"tC":38.7, "tF":101.6}}
```

## Flattened response

```text
aenergy.by_minute[0]=3403.476
aenergy.by_minute[1]=3403.476
aenergy.by_minute[2]=3190.759
aenergy.minute_ts=1773616260
aenergy.total=52565.631
apower=182.8
current=0.841
freq=50
id=0
output=true
ret_aenergy.by_minute[0]=0
ret_aenergy.by_minute[1]=0
ret_aenergy.by_minute[2]=0
ret_aenergy.minute_ts=1773616260
ret_aenergy.total=0
source=init
temperature.tC=38.6
temperature.tF=101.5
voltage=244.8
```
