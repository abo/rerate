import React from 'react';
import ReactDOM from 'react-dom';
import { Sparklines, SparklinesLine, SparklinesBars } from 'react-sparklines';

class Rerate extends React.Component {
    constructor(props) {
        super(props);
        this.state = { histogram:[] };
    }

    componentDidMount() {
        this.intervalId = setInterval( () => {
            fetch("/histogram/"+this.props.counterId)
            .then(response => response.json())
            .then(json => this.setState({histogram:json}));
        },this.props.interval);
    }

    componentWillUnmount() {
        clearInterval(this.intervalId);
    }

    render() {
        const data = this.state.histogram.reverse();
        return (<div>{React.Children.map(this.props.children, function(child) {
                    return React.cloneElement(child, { data });
                })}</div>);
    }
}


ReactDOM.render(
    <Rerate counterId="0" interval="500"><Sparklines>
        <SparklinesLine />
    </Sparklines></Rerate>, document.getElementById("sparkline-0")
);
ReactDOM.render(
    <Rerate counterId="1" interval="500"><Sparklines>
        <SparklinesBars />
    </Sparklines></Rerate>, document.getElementById("sparkline-1")
);
/*ReactDOM.render(
    <Rerate counterId="2" interval="1000"><Sparklines>
        <SparklinesLine />
    </Sparklines></Rerate>, document.getElementById("sparkline-2")
);
ReactDOM.render(
    <Rerate counterId="3" interval="500"><Sparklines>
        <SparklinesLine />
    </Sparklines></Rerate>, document.getElementById("sparkline-3")
);
ReactDOM.render(
    <Rerate counterId="4" interval="1000"><Sparklines>
        <SparklinesLine />
    </Sparklines></Rerate>, document.getElementById("sparkline-4")
);
ReactDOM.render(
    <Rerate counterId="5" interval="1000"><Sparklines>
        <SparklinesLine />
    </Sparklines></Rerate>, document.getElementById("sparkline-5")
 );*/