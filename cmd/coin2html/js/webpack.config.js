const path = require("path");
const HtmlBundlerPlugin = require("html-bundler-webpack-plugin");

module.exports = {
  mode: "development",
  devtool: "inline-source-map",
  output: {
    path: path.join(__dirname, "dist/"),
  },
  plugins: [
    new HtmlBundlerPlugin({
      entry: {
        head: "./head.html",
        body: "./body.html",
      },
      css: {
        inline: true, // inline CSS into HTML
      },
      js: {
        inline: true, // inline JS into HTML
      },
    }),
  ],
  module: {
    rules: [
      {
        test: /\.ts$/,
        use: "ts-loader",
        exclude: /node_modules/,
      },
      {
        test: /\.(css)$/,
        use: ["css-loader"],
      },
      // inline all assets: images, svg, fonts
      {
        test: /\.(png|jpe?g|webp|svg|woff2?)$/i,
        type: "asset/inline",
      },
    ],
  },
  resolve: {
    extensions: [".ts", ".js"],
  },
  performance: false, // disable warning max size
};
