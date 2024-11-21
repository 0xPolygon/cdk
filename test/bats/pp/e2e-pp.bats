setup() {
    load '../../helpers/common-setup'
    _common_setup
}

@test "Verify batches" {
    echo "Waiting 10 minutes to get some settle certificate...."
    run $PROJECT_ROOT/../scripts/agglayer_certificates_monitor.sh 1 600
    assert_success
}
