const path = require('path');
module.exports = {
    entry: './main.jsx',
    output: {
        path: path.resolve('assets'),
        filename: 'sparkline.js'
    },
    module: {
        loaders: [
            {test: /\.jsx$/, loader: 'babel-loader', exclude: /node_modules/}
        ]
    }
}