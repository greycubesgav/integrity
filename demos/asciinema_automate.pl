#!/usr/bin/env perl
use strict;
use warnings;
use IPC::Open3;
use Symbol 'gensym';
use IO::Select;
use File::Basename;

#----------------------------------------------------------------------------------
# Create a dummy STDERR filehandle
my $stderr = gensym;

#----------------------------------------------------------------------------------
# Check if an argument is provided
die "Usage: $0 <input string>\n" unless defined $ARGV[0];

my $command_file=$ARGV[0];
my $command_name=basename($command_file);

my $cast_file = 'demos/output/' . $command_name . ".cast";

my $cols=85;
my $rows=8;

# Get the first argument as the input file name
open my $fh, '<', $command_file or die "Could not open file '$command_file' $!\n";
#my $input = do { local $/; <$fh> };
my @input = <$fh>;
close $fh;

for my $line (@input) {

	print "Read In: $line";
	chomp $line;

	if ($line =~ /^#\scols=(\d+)/) {
		$cols=$1;
		#print "Found line cols=$cols\n";
		next;
	}

	if ($line =~ /^#\srows=(\d+)/) {
		$rows=$1;
		#print "found line rows=$rows\n";
		next;
	}

	my $prompt = '$ ';
	# Bash script parts
	my $bash_cmd = 'bash -c \'' . "printf \"%s\" \"$prompt\"; " . 'for i in ';

	# Loop through each character in the string
	for my $char (split //, $line) {
			# Handle special characters
			if ($char eq "'") {
					$char = "\\'";
			} elsif ($char eq '"') {
					$char = '\\"';
			}
			# Add each character to the bash command
			$bash_cmd .= "\"$char\" ";
	}

	# Complete the bash command with random delays
	$bash_cmd .= '; do printf "%s" "$i"; sleep $(awk "BEGIN { print 0.05 + rand() * (0.075) }"); done; echo;' . "$line";
	#$bash_cmd .= '; sleep 1.5 ';
	$bash_cmd .= "\'";
	print "# Sending command: $bash_cmd\n";

	# Open the process using open3
	my ($stdin, $stdout);
	my $pid = IPC::Open3::open3($stdin, $stdout, $stderr, "asciinema", "rec", "--cols", $cols, '--rows', $rows, '-c', $bash_cmd, '--append', $cast_file);
	#my $pid = IPC::Open3::open3($stdin, $stdout, $stderr, "asciinema", "rec", "--cols", '120', '--rows', '120', '--stdin', '-c', 'bash', '--overwrite', "demo.cast");

	#sleep 3;
	# Send Ctrl-D to asciinema's stdin to signal EOF
	print $stdin "\x04";
	close $stdin; # Close stdin to indicate no more input

	# Wait for the process to finish and collect the exit status
	waitpid($pid, 0);

	# Print the response from asciinema
	#print $response;

}