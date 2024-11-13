setup() {
    load '../../helpers/common-setup'
    _common_setup
}

@test "Verify batches" {
    echo "Waiting 10 minutes to get some verified batch...."
    run $PROJECT_ROOT/test/scripts/batch_verification_monitor.sh 0 600
    assert_success
}
