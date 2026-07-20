use IO::Socket::UNIX;

my $socket = IO::Socket::UNIX->new(
    Type => SOCK_STREAM,
    Peer => '/var/run/docker.sock'
) or die "Can't connect to docker socket: $!";

# Send Container Prune request
print $socket "POST /v1.41/containers/prune?force=true HTTP/1.0\r\n\r\n";
# Read response
while (my $line = <$socket>) {
    print $line;
    last if $line =~ /^\r?\n$/; # end of headers
}
my $body;
while (my $line = <$socket>) {
    $body .= $line;
}
print "Container Prune Body: $body\n";

# Reconnect and Send Image Prune request
$socket = IO::Socket::UNIX->new(
    Type => SOCK_STREAM,
    Peer => '/var/run/docker.sock'
) or die $!;
print $socket "POST /v1.41/images/prune?all=true HTTP/1.0\r\n\r\n";
while (my $line = <$socket>) {
    print $line;
    last if $line =~ /^\r?\n$/;
}
$body = "";
while (my $line = <$socket>) {
    $body .= $line;
}
print "Image Prune Body: $body\n";

# Reconnect and Send Volume Prune request
$socket = IO::Socket::UNIX->new(
    Type => SOCK_STREAM,
    Peer => '/var/run/docker.sock'
) or die $!;
print $socket "POST /v1.41/volumes/prune HTTP/1.0\r\n\r\n";
while (my $line = <$socket>) {
    print $line;
    last if $line =~ /^\r?\n$/;
}
$body = "";
while (my $line = <$socket>) {
    $body .= $line;
}
print "Volume Prune Body: $body\n";
