var path = require('path');
var HtmlWebpackPlugin = require('html-webpack-plugin')
const VueLoaderPlugin = require('vue-loader/lib/plugin')
var distPath = path.resolve(__dirname, 'dist')

module.exports = {
  mode: 'development',
  entry: './main.js',
  output: {
    path: distPath,
    filename: 'wing.bundle.js',
    publicPath: '/'
  },
  module: {
    rules: [
      {
        test: /\.html$/,
        use: ["html-loader"]
      },
      {
        test: /\.js/,
        use: ["babel-loader"],
        exclude: /node_modules/
      },
      {
        test: /\.vue/,
        use: ["vue-loader"]
      },
      {
        test:/\.(png|jpe?g|gif|svg)(\?.*)?$/,
        loader:'url-loader',
        options:{
          limit:10000,
          name:'img/[name].[ext]?[hash]'
        }
      }, 
      {
        test: /\.(css)$/,
        use: [
          'style-loader',
          'css-loader'
        ]
      },
      {
        test: /\.(scss)$/,
        use: [
          'style-loader',
          'css-loader',
          'sass-loader'
        ]
      },
      {
        test: /\.(ttf|ttc|woff(2)?)/,
        loader: 'file-loader',
        options: {
          name: "res/[name].[ext]?[hash]"
        }
      }
    ]
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: "./index.html"
    }),
    new VueLoaderPlugin()
  ],
  devServer: {
    contentBase: distPath,
    compress: true,
    historyApiFallback: {
      rewrites: [
        { from: /^(.*)$/, to: '/' },
      ]
    },
    port: 9000,
    proxy: {
      '/api': "http://localhost:8077/"
    }
  },
  resolve: {
    alias: {
      'vue$': 'vue/dist/vue.esm.js' // 'vue/dist/vue.common.js' for webpack 1
    }
  }
};
