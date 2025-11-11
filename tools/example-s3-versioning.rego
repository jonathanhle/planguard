package accurics

# Example Terrascan policy for testing the converter
# This checks if S3 buckets have versioning enabled

s3Versioning[retVal] {
    bucket := input.aws_s3_bucket[_]
    not bucket.config.versioning[_].enabled == true

    retVal := {
        "Id": bucket.id,
        "ReplaceType": "add",
        "CodeType": "resource",
        "Traverse": sprintf("aws_s3_bucket[%s]", [bucket.name]),
        "Attribute": "versioning.enabled",
        "AttributeDataType": "bool",
        "Expected": true,
        "Actual": false
    }
}
