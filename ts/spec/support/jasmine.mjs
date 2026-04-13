export default {
  spec_dir: ".",
  spec_files: [
    "entities/**/*.spec.ts",
    "indicators/**/!(*.d).spec.ts"
  ],
  helpers: [],
  env: {
    stopSpecOnExpectationFailure: false,
    random: false,
    forbidDuplicateNames: true
  }
}
