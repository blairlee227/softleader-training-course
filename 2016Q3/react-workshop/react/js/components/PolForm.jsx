import React from "react";
import PolStore from "../stores/PolStore.jsx";
import PolAction from "../actions/PolAction.jsx";
import es6BindAll from "es6bindall";

export default class PolForm extends React.Component {

  constructor(props) {
    super(props);
    this.state = {name1: "1"};
    
    es6BindAll(this, [
      "_onChange", "handleInputChange"
    ]);
  }

  componentDidMount() {
    PolStore.addChangeListener(this._onChange);
  }

  componentWillUnmount() {
    PolStore.removeChangeListener(this._onChange);
  }

  _onChange() {
    this.setState(PolStore.getDatas().formData);
  }

  handleInputChange(e) {
    var obj = {};
    obj[e.target.name] = e.target.value;
    this.setState(obj);
  }

  render() {

    return (
      <form>
        <input type="text" name="name1" value={this.state.name1} onChange={this.handleInputChange} /> <br />
      </form>
    );
  }

}