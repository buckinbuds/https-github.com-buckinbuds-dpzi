**MODIFY PRIOR IMPLEMENTATION**
After modification, save file as 'tsconfig.json'.

{
  "compilerOptions": {
    "lib": [
      "es5",
      "es6"
    ],
    "target": "es6",
    "module": "commonjs",
    "moduleResolution": "node",
    "resolveJsonModule": true,
    "esModuleInterop": true,
    "noImplicitAny": true,
    "baseUrl": ".",
    "typeRoots": [
      "node_modules/@types",
      "@types"
    ],
    "outDir": "dist"
  },
  "include": [
    "src/**/*"
  ]
}
