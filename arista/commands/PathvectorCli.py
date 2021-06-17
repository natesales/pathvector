import os

import CliParser
import BasicCli
import TerminalUtil


class PathvectorCmd(object):
    syntax = """pathvector [ <arg> ]"""
    pathvectorArgsRule = CliParser.StringRule("ARG", "Executable arguments and/or switches", 'args')
    data = {'pathvector': 'Run pathvector', '<arg>': pathvectorArgsRule}

    # noinspection PyMethodMayBeStatic
    def pathvector(self, mode, args):
        with TerminalUtil.NoTerminalSettings(mode.session_.terminalCtx_):
            command = "sudo pathvector"
            if args:
                command += " " + args
            os.system(command)

    def handler(self, mode, args):
        self.pathvector(mode, args.get("<arg>", None))


# Register the command by adding a new class
BasicCli.EnableMode.addCommandClass(PathvectorCmd)
