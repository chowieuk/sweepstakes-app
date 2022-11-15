import React from 'react'

const Flags = ({ flags } : {flags : string[]}) => {
    return (
      <>
      {flags.map((flag, index) => (
              <div key={index} className="contentBlock contentBlock--one">
                  <img src={flag} alt=""/>
  
              </div>
          ))} 
      </>
    )
  }
  
  export default Flags