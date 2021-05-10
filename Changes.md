## v0.0.5 2021-05-10

* Remove trailing zeros from parsed package versions because version comparison
  func `version.Compare` does not need trailing zeros to compare versions
  correctly.

## v0.0.4  2020-12-14

* Added support for parsing Ruby package versions, based on the [rubygems
  tests](https://github.com/rubygems/rubygems/blob/master/test/rubygems/test_gem_version.rb).


## v0.0.3  2020-04-23

* Added `.String()` and `.Clone()` methods to the `Version` struct.

* Used [enumer](https://github.com/alvaroloes/enumer) to generate a
  `.String()` method for the `ParsedAs` type.

* Changed the `version.Compare` func so that versions of different length
  compare as equal when both end in one or more segments of zeroes. In other
  words, `1.0` and `1.0.0` are now treated as equal.


## v0.0.2  2020-03-05

* Add the `pkg/name` package for name normalization.


## v0.0.1  2020-03-05

* Initial release
