// import { LitElement } from 'lit';

// const singletons = {};
// let container: HTMLElement;

// export function getSingleton(elName: string) {
//   if (!container) {
//     container = document.querySelector('[singleton-container]') as HTMLElement;
//   }

//   let el = singletons[elName];

//   if (!el) {
//     singletons[elName] = el = document.createElement(elName);
//     container.appendChild(el);
//   }

//   return el;
// }

// export function singleton(name: string) {
//   return (prototype: LitElement, propertyName: string) => {
//     Object.defineProperty(prototype, propertyName, {
//       get() {
//         return getSingleton(name);
//       },
//       enumerable: true,
//       configurable: true,
//     });
//   };
// }
