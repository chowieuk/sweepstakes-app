import React from 'react';
const { useState, useEffect, useRef, useCallback } = React;

const InfiniteLooper = function InfiniteLooper({
  speed,
  direction,
  children,
  animState,
  userTeam,
}: {
  speed: number;
  direction: "right" | "left";
  children: React.ReactNode;
  animState: string;
  userTeam: string;
}) {
  const [looperInstances, setLooperInstances] = useState(1);
  const [looperAnimation, setlooperAnimation] = useState(true);
  const [iterationCount, setIterationCount] = useState<number>(0);
  const outerRef = useRef<HTMLDivElement>(null);
  const innerRef = useRef<HTMLDivElement>(null);
  const instanceRef = useRef<HTMLDivElement>(null);

  let iterCount = 0;

  function resetAnimation() {
    if (innerRef?.current) {
      innerRef.current.setAttribute("data-animate", "false");

      setTimeout(() => {
        if (innerRef?.current) {
          innerRef.current.setAttribute("data-animate", "true");
        }
      }, 10);
    }
  }

  const setupInstances = useCallback(() => {
    if (!innerRef?.current || !outerRef?.current) return;

    const { width } = innerRef.current.getBoundingClientRect();

    const { width: parentWidth } = outerRef.current.getBoundingClientRect();

    const widthDeficit = parentWidth - width;

    const instanceWidth = width / innerRef.current.children.length;

    if (widthDeficit) {
      setLooperInstances(
        looperInstances + Math.ceil(widthDeficit / instanceWidth) + 1
      );
    }

    resetAnimation();
  }, [looperInstances]);
  
  
  /*
    6 instances, 200 each = 1200
    parent = 1700
  */

  useEffect(() => setupInstances(), [setupInstances]);

  // Below resets animation on resize
  // Commenting it out as it will restart the animation when we don't want it to

    //   useEffect(() => {
    //     window.addEventListener("resize", setupInstances);

    //     return () => {
    //       window.removeEventListener("resize", setupInstances);
    //     };
    //   }, [looperInstances, setupInstances]);

  // Adds an event listener to count how many times the animation has looped
  // This will allow us to stop smoothly

    useEffect(() => {
    const handleIteration = () => {
        iterCount++
        setIterationCount(iterCount)
    };

    const element = instanceRef.current;

    element.addEventListener('animationiteration', handleIteration);

    return () => {
        element.removeEventListener('animationiteration', handleIteration);
    };
  });

  return (
    <div className="looper" ref={outerRef}>
      <div className="looper__innerList" ref={innerRef} data-animate={animState} >
        {[...Array(looperInstances)].map((_, ind) => (
          <div
            ref={instanceRef}
            key={ind}
            className="looper__listInstance"
            style={{
              animationDuration: `${speed}s`,
              animationDirection: direction === "right" ? "reverse" : "normal",
              animationIterationCount: looperAnimation ? "infinite" : `${iterationCount + 1}`,
            }}
          >
            {children}
          </div>
        ))}
      </div>
      <div className="reveal-button" onClick={() => {
        setlooperAnimation(false);}
        }>
        {looperAnimation ? "Click here to reveal your team!" : `Your team is ${userTeam}`}!
      </div>
    </div>
  );
}

export default InfiniteLooper