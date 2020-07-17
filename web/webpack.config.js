module.exports = {
  mode: "development",
  entry: "./src/index.js",
  output: {
    filename: "../../static/js/bundle.js",
  },
  devServer: {
    contentBase: "../",
  },
  devtool: 'inline-source-map'
};
