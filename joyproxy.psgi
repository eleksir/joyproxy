use 5.018;
use strict;
use warnings;
use utf8;
use open qw (:std :utf8);
use version; our $VERSION = qw (1.1.0);

use lib qw (. ./vendor_perl/lib/perl5);
use joyproxy qw (joyurl joyproxy);

my $app = sub {
	my $env = shift;

	my $msg = "Your Opinion is very important for us, please stand by.\n";
	my $status = '404';
	my $content = 'text/plain';

	if ($env->{PATH_INFO} =~ /^\/joyproxy\/(.+)/xmsg) {
		my $joyproxyurl = $1;
		($status, $msg) = ('400', "Bad Request?\n");

		if (($joyproxyurl =~ /\.mp4$/xmsgi) or ($joyproxyurl =~ /\.webm$/xmsgi)) {
			($status, $content, $msg) = joyproxy $joyproxyurl ;
		}
	}

	if ($env->{PATH_INFO} =~ /^\/joyurl/xmsg) {
		my $querystring = $env->{QUERY_STRING} // '';

		($status, $content, $msg) = joyurl ($querystring, $env->{HTTP_HOST});
	}

	use bytes;
	my $length = length $msg;
	no bytes;

	return [
		$status,
		[ 'Content-Type' => $content, 'Content-Length' => $length ],
		[ $msg ],
	];
};


__END__
# vim: set ft=perl noet ai ts=4 sw=4 sts=4:
