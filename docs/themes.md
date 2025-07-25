# Custom Themes

If you want `acci-ping` to use custom colours or just want to tweak the defaults this is the place for you.
Themes in `acci-ping` are defined by a simple json schema and can be loaded on application start up with the
`--theme [file]` argument. Note: `acci-ping` does not set a background colour and always uses the current colour.

Here is an example:

```json
{
    "name": "my new theme",
    "version": "1.0.0",
    "colours": {
        "primary": {
            "8-bit": 15
        },
        "secondary": {
            "8-bit": 243
        },
        "highlight": {
            "8-bit": 227
        },
        "emphasis": {
            "8-bit": 87
        },
        "title-highlight": {
            "8-bit": 199
        },
        "positive": {
            "24-bit": "#30FF02"
        },
        "dark-positive": {
            "24-bit": {
                "r": 30,
                "g": 117,
                "b": 11
            }
        },
        "negative": "Red",
        "dark-negative": "DarkRed"
    }
}
```

Example usage:
```sh
acci-ping --theme ~/my-new-theme.json
```

As you can see the format supports many ways of choosing what colour you want. But first an intro into the
keys used:

* `primary` is the main font colour this will be used for the data points and bodies of text.
* `secondary` is the less important font colour for gradients, ideally this should have less contrast than
  `primary`.
* `highlight` is x and y axis label colour.
* `emphasis` is the colour used for URL and date indicators.
* `positive` and `dark-positive` are the "success" or "good outcome" colours (e.g. lowest ping)
* `negative` and `dark-negative` are the "failure" or "bad outcome" colours (e.g. slowest ping)

Next covering how to specify the colours you wish to use. Important note here, the actual colour which will be
displayed is entirely up to the terminal you use and your mileage may vary. There's 3 main ways to specify a
colour:

* `3/4 bit` is specified with no prefix and can only be one of this exact list (these are the most widely
  supported codes and is what the **default** light/dark themes use):
    * `"Black"`
    * `"Gray"`
    * `"LightGray"`
    * `"White"`
    * `"DarkRed"`
    * `"DarkGreen"`
    * `"DarkYellow"`
    * `"DarkBlue"`
    * `"DarkMagenta"`
    * `"DarkCyan"`
    * `"Red"`
    * `"Green"`
    * `"Yellow"`
    * `"Blue"`
    * `"Magenta"`
    * `"Cyan"`
    * `""` is the no colour option

  In the example json `"negative": "Red",` is using the `3/4 bit` colour declaration.
* `8-bit` is the colour according to 8-bit terminal colouring, where each terminal may be implemented
  differently. See this [wikipedia](https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit) link for more
  details, and it's recommend to test these colours in your terminal. To use one of these colours simply open
  a new json object and then the number of the colour you like (between 0 and 255), e.g.:
  ```json
  "primary": {
      "8-bit": 15
  }
  ```

* `24-bit` or "TrueColour" isn't as widely supported so no default themes use them but if your terminal
  supports them then they can be used in a custom theme (note that `acci-ping` doesn't check for support so
  bad output or corruption may occur if you use a theme with these colours and the terminal doesn't support
  them). These are more traditional 8-bits per red, green and blue channel and therefore work like the rest of
  the web and can be specified using traditional (`#11FFCC`) hex codes. Or if you prefer decimal component
  encoding:
  ```json
  "positive": {
      "24-bit": "#30FF02"
  },
  "dark-positive": {
      "24-bit": {
          "r": 30,
          "g": 117,
          "b": 11
      }
  },
  ```


## Disable Colouring

Use the builtin `--theme no-theme` or `--theme no` which doesn't require a custom theme and ensures all text
is uncoloured.
