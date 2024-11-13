setup() {
    load '../../helpers/common-setup'
    _common_setup
}

@test "Verify batches" {
    echo "Waiting 10 minutes to get some settle certificate...."
    run $PROJECT_ROOT/test/scripts/batch_verification_monitor.sh 1 600
    assert_success
}
