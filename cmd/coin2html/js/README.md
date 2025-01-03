This is a Typescript project powering the coin2html single page viewer.

The overarching goal is to have everything packaged in a single html file, the data, the style sheet, the code everything. It needs to run in the browser without being backed by any server infrastructure, just open the HTML file and that's it.

For the longest time I tried to avoid using a JS builder, just piecing the final HTML file together from the various bits here. However the un-modularized Typescript code was difficult to test. I had hopes for ES module support in the browser, but it turns out that inlined modules are useless, cannot be imported (without resorting to some ugly hacks), so eventually I gave up and introduced webpack. I picked webpack because of its HtmlBundlerPlugin that can produce a single HTML page bundling everything.

D3 is the main workhorse of the app, it builds all the visual elements beyond the bare bones dom skeleton defined in body.html. That include not just the charts but the html tables as well. I'd like to stick with D3 as much as possible to keep dependencies to a minimum. I might be tempted to switch to Observable/Plot to reduce the amount of code given that at least some of the charts are pretty standard stock, we'll see, at the moment I'm having too much fun with D3.

As it is the final html file for examples/yearly is ~7MB. This is after stopping importing D3 wholesale and importing just the bits that are used. It does include sourcemaps though to make debugging convenient. There's no minification yet either, which would presumably help as well.

# Project structure

The main pieces are

- head.html - defines the HTML head part and pulls in styles.css
- body.html - lays out the basic/static page structure and pulls in the typescript modules
- styles.css - minimal styling for the page elements
- src/ - contains all the typescript modules

The bundler plugin pulls all of the above together producing two files in dist/

- body.html
- head.html

coin2html looks for these two files to combine them with the JSON data and produce the final HTML page. This is all orchestrated by the `webpack.config.js`. The `npm build` script executes the process.

# Testing

`jest` is used as its the defacto standard these days. Use `npm test` to run the test.
