package joyproxy;

use 5.018;
use strict;
use warnings "all";
use utf8;
use open qw(:std :utf8);

use HTTP::Tiny;
use URI::URL;
use URI::Escape qw(uri_unescape);

use Exporter qw(import);
use vars qw/$VERSION/;

$VERSION = "1.0";
our @EXPORT = qw(joyproxy joyurl);

$| = 1;

sub joyproxy ($) {
	my $str = shift;
	chomp($str);

	return ('500', 'text/plain', 'This is not reactor video') if ($str !~ /^img[0-9]+\.reactor\.cc/);

	my ($file, $filesize) = __dlfunc("http://$str");

	if (defined($file)) {
		my $buf = '';

		if (open(VID, '<', $file)) {
			read(VID, $buf, $filesize);
			close VID;

			if ($file =~ /\.mp4/i) {
				return ('200', 'video/mp4', $buf);
			} elsif ($file =~ /\.webm/i) {
				return ('200', 'video/webm', $buf);
			} else {
				return ('200', 'video/mpeg', $buf);
			}
		} else {
			unlink $file if (-f $file);
			return ('500', 'text/plain', "Unable to open file in temporary loaction: $!\n");
		}
	} else {
		return ('500', 'text/plain', "Unable to get file from remote source!\n");
	}
}

sub joyurl (@) {
	my $str = shift;
	my $host = shift;
	my $prefix = shift;
	$str = uri_unescape($str);

	if (length($str) < 60) {
		$str = '';
	} else {
		$str = substr($str, 14);
	}

	if ($str =~ /^img\d+\.reactor\.cc/) {
		# img1.reactor.cc/pics/post/webm/видосик.webm
		my @url = split(/\//, $str);

		if (($url[3] eq 'webm' || $url[3] eq 'mp4') and
		    ($url[4] =~ /\.webm$/ || $url[4] =~ /\.mp4$/)) {
			# we prefer mp4, right?, so
			my $fname;
			$fname = substr($url[4], 0, -4) if (substr($url[4], -4, 4) eq '.mp4');
			$fname = substr($url[4], 0, -5) if (substr($url[4], -5, 6) eq '.webm');

			$str = sprintf(
				"https://%s/%s/joyproxy/%s/%s/%s/mp4/%s.mp4",
				$host,
				$prefix,
				$url[0],
				$url[1],
				$url[2],
				$fname
			);

			$str = __urlencode($str);
		} else {
			$str = '';
		}
	} else {
		$str = '';
	}

	my $msg = "<html>\n<body>\n<form method='get' action='$prefix/joyurl'>\n<input type='text' name='joyurl' size=100 autofocus><br />\n<input type='submit' value='Post it!'' style='font-size:115%;'' />\n<br>$str\n</body>\n</html>\n";

	return ('200', 'text/html', $msg);
}

sub __urlencode($) {
	my $str = shift;
	my $urlobj = url $str;
	$str = $urlobj->as_string;
	$urlobj = undef;
	undef $urlobj;
	return $str;
}

sub __dlfunc($) {
	my $url = shift;

	if (! -d "/tmp/joyproxy") {
		warn "no /tmp/joyproxy";
		return (undef, undef) unless (mkdir("/tmp/joyproxy"));
	}

	my @tmparray = split(/\//, $url);
	my $file = $tmparray[@tmparray - 1];
	$#tmparray = -1; undef @tmparray;
	$file = "/tmp/joyproxy/" . $file;
	$url = __urlencode($url);

	my $http = HTTP::Tiny->new(max_size => 10485760);

	my $response = $http->get(
		$url,
		{
			headers => {
				"Accept" => '*/*',
				"Accept-Encoding" => 'identity;q=1, *;q=0',
				"Accept-Language" => 'ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7',
				"Range" => 'bytes=0-',
				"Referer" => 'http://old.reactor.cc/all',
				"User-Agent" => "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.116 Safari/537.36"
			}
		}
	);

	if ($response->{success}) {
		my $filesize = 0;

		open (FILE, '>', $file) or do {
			$http = undef;
			$response = undef;
			return (undef, undef);
		};

		binmode FILE;

		print (FILE $response->{content}) or do {
			$http = undef;
			$response = undef;
			return (undef, undef);
		};

		close FILE;
		$http = undef;
		$response = undef;
		$filesize = (stat($file))[7];
		return ($file, $filesize);
	} else {
		$http = undef;
		$response = undef;
		return (undef, undef);
	}
}


1;

# vim: set ft=perl noet ai ts=4 sw=4 sts=4:
