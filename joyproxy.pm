package joyproxy;

use 5.018;
use strict;
use warnings;
use utf8;
use open qw (:std :utf8);

use Fcntl;
use HTTP::Tiny;
use URI::URL;
use URI::Escape qw (uri_unescape);

use Exporter qw (import);
use version; our $VERSION = qw (1.1.0);
our @EXPORT_OK = qw (joyproxy joyurl);

sub joyproxy ($) {
	my $str = shift;
	chomp $str;

	# formal url check
	if ($str !~ /^img\d+\.reactor\.cc/xmsg) {
		return ('500', 'text/plain', 'This is not reactor video');
	}

	my ($file, $filesize) = __dlfunc ("http://$str");

	if (defined $file) {
		my $buf = '';

		if (open my $VID, '<', $file) {
			binmode $VID;
			my $readlen = read ($VID, $buf, $filesize);
			close $VID;

			if (defined $readlen) {
				if ($readlen != $filesize) {
					carp "Unable to read $file readlen does not match filesize: $readlen vs $filesize";
					return ('500', 'text/plain', "Unable to read file in temporary location\n");
				}
			} else {
				carp "Unable to read $file: $!";
				return ('500', 'text/plain', "Unable to read file in temporary location: $!\n");
			}

			if ($file =~ /\.mp4$/xmsgi) {
				return ('200', 'video/mp4', $buf);
			} elsif ($file =~ /\.webm$/xmsgi) {
				return ('200', 'video/webm', $buf);
			} else {
				return ('200', 'video/mpeg', $buf);
			}
		} else {
			if (-f $file) {
				unlink $file;
			}

			return ('500', 'text/plain', "Unable to open file in temporary location: $!\n");
		}
	} else {
		return ('500', 'text/plain', "Unable to get file from remote source!\n");
	}
}

sub joyurl (@) {
	my $str = shift;
	my $host = shift;

	# TODO: handle error here
	$str = uri_unescape $str;

	if (length ($str) < 60) {
		$str = '';
	} else {
		$str = substr $str, 14;
	}

	if ($str =~ /^img\d+\.reactor\.cc/xmsg) {
		# img1.reactor.cc/pics/post/webm/видосик.webm
		my @url = split /\//msx, $str;

		if (($url[3] eq 'webm' || $url[3] eq 'mp4') &&
		    ($url[4] =~ /\.webm$/xmsg || $url[4] =~ /\.mp4$/xmsg)) {
			# we prefer mp4, right?, so
			my $fname;
			$fname = substr ($url[4], 0, -5) if (substr ($url[4], -5, 6) eq '.webm');
			$fname = substr ($url[4], 0, -4) if (substr ($url[4], -4, 4) eq '.mp4');

			$str = sprintf (
				'https://%s/joyproxy/%s/%s/%s/mp4/%s.mp4',
				$host,
				$url[0],
				$url[1],
				$url[2],
				$fname
			);

			$str = __urlencode ($str);
		} else {
			$str = '';
		}
	} else {
		$str = '';
	}

	my $msg = << "EOL";
<html>
<body>
<form method='get' action='joyurl'>
<input type='text' name='joyurl' size=100 autofocus><br />
<input type='submit' value='Post it!' style='font-size:115%;' />
<br>$str
</body>
</html>\n
EOL

	return ('200', 'text/html', $msg);
}

sub __urlencode($) {
	my $str = shift;
	my $urlobj = url $str;
	$str = $urlobj->as_string;
	$urlobj = '';
	return $str;
}

sub __dlfunc($) {
	my $url = shift;

	unless (-d '/tmp/joyproxy') {
		carp 'no /tmp/joyproxy';

		unless (mkdir '/tmp/joyproxy') {
			return (undef, undef);
		}
	}

	my @explodedurl = split /\//msx, $url;
	my $file = $explodedurl[$#explodedurl]; ## no critic (Variables::RequireNegativeIndices)
	$#explodedurl = -1; undef @explodedurl;
	$file = '/tmp/joyproxy/' . $file;
	$url = __urlencode ($url);

	my $http = HTTP::Tiny->new (max_size => 10485760);

	# use ->get because of ->mirror does not respect max_size
	my $response = $http->get (
		$url,
		{
			headers => {
				'Accept' => '*/*',
				'Accept-Encoding' => 'identity;q=1, *;q=0',
				'Accept-Language' => 'ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7',
				'Range' => 'bytes=0-',
				'Referer' => 'http://old.reactor.cc/all',
				'User-Agent' => 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.116 Safari/537.36'
			}
		}
	);

	$http = '';

	if ($response->{success}) {
		my $filesize = 0;

		sysopen (my $FILE, $file, O_CREAT|O_TRUNC|O_WRONLY) or do {
			$response = '';
			return (undef, undef);
		};

		binmode $FILE;

		syswrite ($FILE, $response->{content}) or do {
			close $FILE;
			$response = '';
			return (undef, undef);
		};

		close $FILE;
		$response = '';
		$filesize = (stat $file )[7];
		return ($file, $filesize);
	} else {
		$response = '';
		return (undef, undef);
	}
}

1;

# vim: set ft=perl noet ai ts=4 sw=4 sts=4:
