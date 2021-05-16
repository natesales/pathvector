import os

import CliParser
import BasicCli
import TerminalUtil


class Bgpq4Cmd(object):
    syntax = """bgpq4 [ <arg> ]"""
    bgpq4ArgsRule = CliParser.StringRule("ARG", "Executable arguments and/or switches", 'args')
    data = {'bgpq4': 'Run bgpq4 IRR parser', '<arg>': bgpq4ArgsRule}

    # noinspection PyMethodMayBeStatic
    def bgpq4(self, mode, args):
        with TerminalUtil.NoTerminalSettings(mode.session_.terminalCtx_):
            command = "bgpq4"
            if args:
                command += " " + args
            os.system(command)

    def handler(self, mode, args):
        self.bgpq4(mode, args.get("<arg>", None))


# Register the command by adding a new class
BasicCli.EnableMode.addCommandClass(Bgpq4Cmd)
