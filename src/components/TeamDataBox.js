import React from "react";

export default function TeamDataBox(props) {
  return (
    <div>
      <div className="databox-header">{props.header}</div>
      <div className="databox-value">{props.value}</div>
    </div>
  );
}
