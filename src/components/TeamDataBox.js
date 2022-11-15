import React from "react";

export default function TeamDataBox(props) {
  return (
    <div className="teamdata-box">
      <div className="databox-header">{props.header}</div>
      <div className="databox-value">{props.value}</div>
    </div>
  );
}
