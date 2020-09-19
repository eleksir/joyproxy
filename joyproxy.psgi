use 5.018;
use strict;
use warnings "all";
use utf8;
use open qw(:std :utf8);

use lib qw(. ./vendor_perl/lib/perl5);
use joyproxy qw(joyurl joyproxy);

$| = 1;

my $app = sub {
	my $env = shift;

	my $msg = "Your Opinion is very important for us, please stand by.\n";
	my $status = '404';
	my $content = 'text/plain';

	if ($env->{PATH_INFO} =~ /^\/joyproxy\/(.+)/) {
		my $joyproxyurl = $1;
		($status, $content, $msg) = ('400', $content, "Bad Request?\n");

		if (($joyproxyurl =~ /\.mp4/i) or ($joyproxyurl =~ /\.webm/i)) {
			($status, $content, $msg) = joyproxy($joyproxyurl);
		}
	}

	if ($env->{PATH_INFO} =~ /^\/joyurl/) {
		if (defined($env->{QUERY_STRING})) {
			($status, $content, $msg) = joyurl($env->{QUERY_STRING}, $env->{HTTP_HOST});
		} else {
			($status, $content, $msg) = joyurl('', $env->{HTTP_HOST});
		}
	}

	use bytes;
	my $length = length($msg);
	no bytes;

	return [
		$status,
		[ 'Content-Type' => $content, 'Content-Length' => $length ],
		[ $msg ],
	];
};


__END__
# vim: set ft=perl noet ai ts=4 sw=4 sts=4:
