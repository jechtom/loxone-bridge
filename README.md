# LoxoneApiBridge

## Úvod a motivace
**LoxoneBridge** je jednoduchá služba, jejíž hlavním cílem je doplnit a rozšířit síťové integrační možnosti systému Loxone Miniserver tam, kde pokulhávají.

## Hlavní funkce

1. **Podpora Digest Authentication**
   Translací (překladem) Loxone kompatibilní Basic Authentication do Digest Authentication pro komunikaci se zařízeními třetích stran, která vyžadují složitější způsob ověřování. Například vhodné pro výrobky, které nepodporují Basic Auth jako výrobky od Shelly nebo Dahua.
2. **Parsování JSON (JSON Path)**
   Přijímá komplexní JSON odpovědi a převede do formátu, který lze snadno a jednoznačně parsovat v Loxone.
3. **Příjem HTTP hooků a překlad na UDP**
   Umí přijímat HTTP callbacky (webhooky) z jiných systémů a stavové změny plynule přeposílat do Loxone miniserveru jako UDP.

## Filozofie a konfigurace

Aplikace je navržena pro maximální jednoduchost provozu. Je bezstavová a není potřeba definovat konfigurační soubory. **Veškerá konfigurace je obsažena a deklarována přímo v URL adrese.** Loxone odesílá své dotazy s instrukcemi ohledně typu překladu přímo v cestě (path) požadavku.

### Příklady použití (Routování přes URL)

* **Překlad Basic Auth na Digest Auth (HTTP)**
  `GET /digest/http/192.168.1.10/aaa`
  Převezme Basic autentizaci zaslanou Loxonem, naváže s cílem komunikaci, zpracuje případný 401 Unauthorized challenge a vyřídí dotaz proti `http://192.168.1.10/aaa` přes Digest Auth.

* **Ignorování chyb HTTPS certifikátu**
  `GET /https-ignore-cert/192.168.1.10/aaa`
  Směruje na HTTPS adresu `https://192.168.1.10/aaa` a záměrně ignoruje chyby neplatného nebo self-signed certifikátu u koncového zařízení.

* **Odesílání UDP paketů přes HTTP**
  `GET /udp/192.168.1.10:444/data`
  Záchytný bod pro odeslání UDP. Jakékoliv odeslané tělo (request body) je aplikací LoxoneApiBridge vzato a beze změny vypáleno jako raw UDP datagram na IP `192.168.1.10` a port `444`.

* **Přeloží JSON jako plochý seznam hodnot**
  `GET /flatten-json/http/192.168.1.10/aaaa`
  Navštíví adresu `http://192.168.1.10/aaaa` a vrací JSON v plochém formátu, například:
  ```json
  {
    "data": {
        "volume": 124,
        "error": false
    },
    "name": "device-1",
    "versions": [ "a", "b", "c" ]
  }
  ```
  Konvertuje na:
  ```
  data.volume=124
  data.error=false
  name=device-1
  versions[0]=a
  versions[1]=b
  versions[2]=c
  ```

## Spuštění

Službu lze nastartovat řadou způsobů. Například pomocí docker nebo vlastního sestavení. 

TODO: Informace doplním po dokončení.

# Dokumentace

## Skladba adresy požadavků

Modifikátory se přidávají na začátek adresy.

Formát:
```
http://loxone-bridge/{modifiers}/{protocol}/{address}/{path-and-query}
       |-----------| |---------| |------------------| |--------------|
       |             |           |                    |
       |             |           |                    +- cesta, která se přeposílá
       |             |           |
       |             |           +- protokol a adresa serveru, kam požadavek poslat
       |             |
       |             +- nula nebo více modifikátorů
       |
       +- adresa Loxone Bridge
      
```

## Modifiers

### Konverze Basic Auth na Digest

Segment: `/digest`

Konvertuje basic auth na digest auth. Basic auth se v Loxonu přidává před adresu serveru. Celková adresa zadaná v loxonu může být například:
```
http://admin:PASSWORD@loxone-bridge/digest/http/10.0.0.5/cgi-bin/accessControl.cgi?action=openDoor&channel=1
```
Bridge vytvoří HTTP request s DIGEST autentizací na adresu:
```
http://10.0.0.5/cgi-bin/accessControl.cgi?action=openDoor&channel=1
```

## Protocols

### HTTP

Segment: `/http`

### HTTPS

Segment: `/https`

### HTTPS + ignore certificate errors

Segment: `/https-ignore-cert`

### UDP

Segment: `/udp`

Pokud se použije, odešle UDP data na adresu. 

Při použití tohoto protokolu se vše v segmentu `{path-and-query}` odešle jako obsah UDP data paketu.

## Address

Adresa, kam se odešle požadavek (`/adresa`). IP adresa nebo DNS jméno. 

Lze uvést i port (`/adresa:port`). Port je u UDP povinný, u HTTP a HTTPS se použije výchozí (80/443).